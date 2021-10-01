.PHONY: deploy

deploy:
	python main.py --runner DataflowRunner \
				   --job_name "compute-`date +"%Y%m%d"`-`uuid  | cut -d- -f1`" \
				   --project paul-test-321412 --region europe-west2
